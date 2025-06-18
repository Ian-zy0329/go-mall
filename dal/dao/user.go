package dao

import (
	"context"
	"errors"
	"github.com/Ian-zy0329/go-mall/common/enum"
	"github.com/Ian-zy0329/go-mall/common/errcode"
	"github.com/Ian-zy0329/go-mall/common/util"
	"github.com/Ian-zy0329/go-mall/dal/model"
	"github.com/Ian-zy0329/go-mall/logic/do"
	"gorm.io/gorm"
)

type UserDao struct {
	ctx context.Context
}

func NewUserDao(ctx context.Context) *UserDao {
	return &UserDao{
		ctx: ctx,
	}
}

func (ud *UserDao) CreateUser(userInfo *do.UserBaseInfo, userPasswordHash string) (*model.User, error) {
	userModel := new(model.User)
	err := util.CopyProperties(userModel, userInfo)
	if err != nil {
		err = errcode.Wrap("UserDaoCreaeteUserError", err)
		return nil, err
	}
	userModel.Password = userPasswordHash
	err = DBMaster().WithContext(ud.ctx).Create(userModel).Error
	if err != nil {
		err = errcode.Wrap("UserDaoCreaeteUserError", err)
		return nil, err
	}
	return userModel, nil
}

func (ud *UserDao) FindUserByLoginName(loginName string) (*model.User, error) {
	user := new(model.User)
	err := DB().Where(model.User{LoginName: loginName}).First(&user).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return user, nil
}

func (ud *UserDao) FindUserById(userId int64) (*model.User, error) {
	user := new(model.User)
	err := DB().Where(model.User{ID: userId}).Find(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (ud *UserDao) UpdateUser(user *model.User) error {
	err := DB().Model(user).Updates(user).Error
	return err
}

func (ud *UserDao) CreateUserAddress(address *do.UserAddressInfo) (*model.UserAddress, error) {
	addressModel := new(model.UserAddress)
	err := util.CopyProperties(addressModel, address)
	if err != nil {
		err = errcode.Wrap("UserDaoCreateUserAddressError", err)
		return nil, err
	}
	var defaultAddress *model.UserAddress
	if addressModel.Default == enum.AddressIsUserDefault {
		defaultAddress, err = ud.GetUserDefaultAddress(addressModel.UserId)
		if err != nil {
			return nil, err
		}
	}

	if defaultAddress != nil && defaultAddress.ID != 0 {
		err = DBMaster().Transaction(func(tx *gorm.DB) error {
			err := tx.WithContext(ud.ctx).Model(defaultAddress).Select("Default").
				Updates(model.UserAddress{Default: enum.AddressIsNotUserDefault}).Error
			if err != nil {
				return err
			}
			err = tx.WithContext(ud.ctx).Create(addressModel).Error
			return err
		})
	} else {
		err = DBMaster().WithContext(ud.ctx).Create(addressModel).Error
	}
	if err != nil {
		return nil, err
	}
	return addressModel, nil
}

func (ud *UserDao) GetUserDefaultAddress(userId int64) (*model.UserAddress, error) {
	address := new(model.UserAddress)
	err := DB().WithContext(ud.ctx).Where(model.UserAddress{UserId: userId, Default: enum.AddressIsUserDefault}).First(&address).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return address, nil
}

func (ud *UserDao) GetUserAddresses(userId int64) ([]*model.UserAddress, error) {
	addresses := make([]*model.UserAddress, 0)
	err := DB().WithContext(ud.ctx).Where(model.UserAddress{UserId: userId}).Find(&addresses).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return addresses, nil
}

func (ud *UserDao) UpdateUserAddress(address *do.UserAddressInfo) error {
	addressModel := new(model.UserAddress)
	err := util.CopyProperties(addressModel, address)
	if err != nil {
		err = errcode.Wrap("UpdateUserAddressError", err)
		return err
	}
	var defaultAddress *model.UserAddress
	if address.Default == enum.AddressIsUserDefault {
		defaultAddress, err = ud.GetUserDefaultAddress(address.UserId)
		if err != nil {
			return err
		}
	}
	if defaultAddress != nil && defaultAddress.ID != 0 && defaultAddress.ID != address.ID {
		err = DBMaster().Transaction(func(tx *gorm.DB) error {
			err := tx.WithContext(ud.ctx).Model(defaultAddress).Select("default").Updates(&model.UserAddress{Default: enum.AddressIsNotUserDefault}).Error
			if err != nil {
				return err
			}
			err = tx.WithContext(ud.ctx).Model(addressModel).Updates(addressModel).Error
			return err
		})
	} else {
		err = DBMaster().WithContext(ud.ctx).Model(addressModel).Updates(addressModel).Error
	}
	return err
}

func (ud *UserDao) GetSingleAddress(addressId int64) (*model.UserAddress, error) {
	address := new(model.UserAddress)
	err := DB().WithContext(ud.ctx).Where(&model.UserAddress{ID: addressId}).Find(&address).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return address, nil
}

func (ud *UserDao) DeleteOneAddress(address *model.UserAddress) error {
	return DBMaster().Delete(address).Error
}
