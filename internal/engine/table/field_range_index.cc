/**
 * Copyright 2019 The Gamma Authors.
 *
 * This source code is licensed under the Apache License, Version 2.0 license
 * found in the LICENSE file in the root directory of this source tree.
 */

#include "field_range_index.h"

#include <math.h>
#include <roaring/roaring.h>
#include <string.h>

#include <algorithm>
#include <cassert>
#include <cstddef>
#include <cstdint>
#include <ctime>
#include <functional>
#include <iostream>
#include <iterator>
#include <limits>
#include <map>
#include <mutex>
#include <numeric>
#include <sstream>
#include <typeinfo>

#include "util/bitmap.h"
#include "util/log.h"
#include "util/utils.h"

namespace vearch {

class FieldRangeIndex {
 public:
  FieldRangeIndex(int field_idx, enum DataType field_type, std::string &name);
  ~FieldRangeIndex();

  bool IsNumeric() { return is_numeric_; }

  enum DataType DataType() { return data_type_; }

  char *Delim() { return kDelim_; }

 private:
  bool is_numeric_;
  enum DataType data_type_;
  char *kDelim_;
  std::string name_;
  long add_num_ = 0;
  long delete_num_ = 0;
};

FieldRangeIndex::FieldRangeIndex(int field_idx, enum DataType field_type,
                                 std::string &name)
    : name_(name) {
  if (field_type == DataType::STRING || field_type == DataType::STRINGARRAY) {
    is_numeric_ = false;
  } else {
    is_numeric_ = true;
  }
  data_type_ = field_type;
  kDelim_ = const_cast<char *>("\001");
}

FieldRangeIndex::~FieldRangeIndex() {}

/**
 * Reverse the byte order of the input array and store the result in the output
 * array. This function also adds 0x80 to the first byte of the output array.
 */
static int ReverseEndian(const unsigned char *in, unsigned char *out,
                         uint len) {
  std::reverse_copy(in, in + len, out);
  out[0] += 0x80;
  return 0;
}

MultiFieldsRangeIndex::MultiFieldsRangeIndex(std::string &path, Table *table,
                                             StorageManager *storage_mgr)
    : path_(path),
      table_(table),
      fields_(table->FieldsNum()),
      storage_mgr_(storage_mgr) {
  std::fill(fields_.begin(), fields_.end(), nullptr);
  field_rw_locks_ = new pthread_rwlock_t[fields_.size()];
  for (size_t i = 0; i < fields_.size(); i++) {
    if (pthread_rwlock_init(&field_rw_locks_[i], nullptr) != 0) {
      LOG(ERROR) << "init lock failed!";
    }
  }
}

MultiFieldsRangeIndex::~MultiFieldsRangeIndex() {
  for (size_t i = 0; i < fields_.size(); i++) {
    if (fields_[i]) {
      pthread_rwlock_wrlock(&field_rw_locks_[i]);
      delete fields_[i];
      fields_[i] = nullptr;
      pthread_rwlock_unlock(&field_rw_locks_[i]);
    }
    pthread_rwlock_destroy(&field_rw_locks_[i]);
  }
  delete[] field_rw_locks_;
}

int MultiFieldsRangeIndex::Delete(int64_t docid, int field) {
  FieldRangeIndex *index = fields_[field];
  if (index == nullptr) {
    return 0;
  }
  std::string value;
  int ret = table_->GetFieldRawValue(docid, field, value);
  if (ret != 0) {
    return ret;
  }

  ret = DeleteDoc(docid, field, value);

  return ret;
}

int MultiFieldsRangeIndex::AddDoc(int64_t docid, int field) {
  if (cf_id_map_.find(field) == cf_id_map_.end()) {
    return 0;
  }

  std::string key;
  int ret = table_->GetFieldRawValue(docid, field, key);
  if (ret != 0) {
    LOG(ERROR) << "get doc " << docid << " failed";
    return ret;
  }
  std::string key_str;

  if (fields_[field]->IsNumeric()) {
    size_t key_len = key.size();
    std::vector<unsigned char> key2(key_len);
    ReverseEndian(reinterpret_cast<const unsigned char *>(key.data()),
                  key2.data(), key_len);
    std::string key2_str(key2.begin(), key2.end());

    key_str = key2_str + ":" + std::to_string(docid);
  } else {
    key_str = key + ":" + std::to_string(docid);
  }

  pthread_rwlock_wrlock(&field_rw_locks_[field]);
  auto &db = storage_mgr_->GetDB();
  int cf_id = cf_id_map_[field];
  rocksdb::ColumnFamilyHandle *cf_handler =
      storage_mgr_->GetColumnFamilyHandle(cf_id);
  rocksdb::Status s = db->Put(rocksdb::WriteOptions(), cf_handler,
                              rocksdb::Slice(key_str), std::to_string(docid));
  if (!s.ok()) {
    std::stringstream msg;
    msg << "rocksdb put error:" << s.ToString() << ", key=" << key_str;
    LOG(ERROR) << msg.str();
    return -1;
  }
  pthread_rwlock_unlock(&field_rw_locks_[field]);

  return 0;
}

int MultiFieldsRangeIndex::DeleteDoc(int64_t docid, int field,
                                     std::string &key) {
  if (cf_id_map_.find(field) == cf_id_map_.end()) {
    LOG(ERROR) << "field index is null, field=" << field;
    return 0;
  }

  pthread_rwlock_wrlock(&field_rw_locks_[field]);
  auto &db = storage_mgr_->GetDB();
  int cf_id = cf_id_map_[field];
  rocksdb::ColumnFamilyHandle *cf_handler =
      storage_mgr_->GetColumnFamilyHandle(cf_id);

  std::string key_str;
  if (fields_[field]->IsNumeric()) {
    size_t key_len = key.size();
    std::vector<unsigned char> key2(key_len);
    ReverseEndian(reinterpret_cast<const unsigned char *>(key.data()),
                  key2.data(), key_len);
    std::string key2_str(key2.begin(), key2.end());

    key_str = key2_str + ":" + std::to_string(docid);
  } else {
    key_str = key + ":" + std::to_string(docid);
  }

  rocksdb::Status s =
      db->Delete(rocksdb::WriteOptions(), cf_handler,
                 rocksdb::Slice(key_str + ":" + std::to_string(docid)));
  if (!s.ok()) {
    std::stringstream msg;
    msg << "rocksdb delete error:" << s.ToString() << ", key=" << key;
    LOG(ERROR) << msg.str();
    return -1;
  }

  pthread_rwlock_unlock(&field_rw_locks_[field]);

  return 0;
}

template <typename Type>
static void AdjustBoundary(std::string &boundary, int offset) {
  static_assert(std::is_fundamental<Type>::value, "Type must be fundamental.");

  if (boundary.size() >= sizeof(Type)) {
    Type b;
    std::vector<char> vec(sizeof(b));
    memcpy(&b, boundary.data(), sizeof(b));
    b += offset;
    memcpy(vec.data(), &b, sizeof(b));
    boundary = std::string(vec.begin(), vec.end());
  }
}

int64_t MultiFieldsRangeIndex::Search(
    const std::vector<FilterInfo> &origin_filters,
    MultiRangeQueryResults *out) {
  out->Clear();

  std::vector<FilterInfo> filters;

  for (const auto &filter : origin_filters) {
    FieldRangeIndex *index = fields_[filter.field];
    if (index == nullptr) {
      LOG(DEBUG) << "field index is null, field=" << filter.field;
      return -1;
    }
    if (not index->IsNumeric() && (filter.is_union == FilterOperator::And)) {
      // type is string and operator is "and", split this filter
      std::vector<std::string> items =
          utils::split(filter.lower_value, index->Delim());
      for (std::string &item : items) {
        FilterInfo f = filter;
        f.lower_value = item;
        filters.push_back(f);
      }
      continue;
    }
    filters.push_back(filter);
  }

  auto fsize = filters.size();

  RangeQueryResult result;
  RangeQueryResult result_not_in;
  bool has_result = false;
  bool has_result_not_in = false;
  result_not_in.SetNotIn(true);
  int64_t retval = 0;

  for (size_t i = 0; i < fsize; ++i) {
    RangeQueryResult tmp_result;
    RangeQueryResult tmp_result_not_in;
    auto &filter = filters[i];
    FieldRangeIndex *index = fields_[filter.field];

    if (not filter.include_lower) {
      if (index->DataType() == DataType::INT) {
        AdjustBoundary<int>(filter.lower_value, 1);
      } else if (index->DataType() == DataType::LONG) {
        AdjustBoundary<long>(filter.lower_value, 1);
      }
    }

    if (not filter.include_upper) {
      if (index->DataType() == DataType::INT) {
        AdjustBoundary<int>(filter.upper_value, -1);
      } else if (index->DataType() == DataType::LONG) {
        AdjustBoundary<long>(filter.upper_value, -1);
      }
    }
    pthread_rwlock_rdlock(&field_rw_locks_[filters[i].field]);

    auto &db = storage_mgr_->GetDB();
    int cf_id = cf_id_map_[filter.field];
    rocksdb::ColumnFamilyHandle *cf_handler =
        storage_mgr_->GetColumnFamilyHandle(cf_id);
    std::string value;
    rocksdb::ReadOptions read_options;
    std::unique_ptr<rocksdb::Iterator> it(
        db->NewIterator(read_options, cf_handler));

    std::string lower_key, upper_key;

    if (index->IsNumeric()) {
      size_t lower_len = filter.lower_value.length();
      size_t upper_len = filter.upper_value.length();
      std::vector<unsigned char> key_l(lower_len);
      std::vector<unsigned char> key_u(upper_len);
      ReverseEndian(
          reinterpret_cast<const unsigned char *>(filter.lower_value.data()),
          key_l.data(), lower_len);
      ReverseEndian(
          reinterpret_cast<const unsigned char *>(filter.upper_value.data()),
          key_u.data(), upper_len);
      lower_key = std::string(key_l.begin(), key_l.end()) + ":";
      upper_key = std::string(key_u.begin(), key_u.end()) + ":";
      for (it->Seek(lower_key); it->Valid(); it->Next()) {
        std::string key = it->key().ToString();
        if (key >= upper_key) {
          break;
        }
        int64_t docid = std::stoll(it->value().ToString());
        if (filter.is_union == FilterOperator::Not) {
          tmp_result_not_in.Add(docid);
          has_result_not_in = true;
        } else {
          result.Add(docid);
          has_result = true;
        }
        retval++;
      }
    } else {
      std::vector<std::string> items =
          utils::split(filter.lower_value, index->Delim());
      for (std::string &item : items) {
        lower_key = item + ":";

        for (it->Seek(lower_key);
             it->Valid() && it->key().starts_with(lower_key); it->Next()) {
          int64_t docid = std::stoll(it->value().ToString());
          if (filter.is_union == FilterOperator::Not) {
            result_not_in.Add(docid);
            has_result_not_in = true;
          } else {
            result.Add(docid);
            has_result = true;
          }
          retval++;
        }
      }
    }
    pthread_rwlock_unlock(&field_rw_locks_[filters[i].field]);
  }

  if (has_result_not_in && has_result) {
    result.IntersectionWithNotIn(result_not_in);
  } else if (has_result_not_in) {
    result = std::move(result_not_in);
  }

  out->Add(std::move(result));
  return retval;
}

int MultiFieldsRangeIndex::AddField(int field, enum DataType field_type,
                                    std::string &field_name) {
  FieldRangeIndex *index = new FieldRangeIndex(field, field_type, field_name);
  fields_[field] = index;
  int cf_id =
      storage_mgr_->CreateColumnFamily("scalar:" + std::to_string(field));
  LOG(INFO) << "Create column family scalar:" << field << " cf_id=" << cf_id;
  cf_id_map_[field] = cf_id;
  return 0;
}

}  // namespace vearch
